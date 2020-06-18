import {
  BeforeInsert,
  BeforeUpdate,
  Column,
  CreateDateColumn,
  Entity,
  OneToMany,
  PrimaryColumn,
  getRepository,
} from 'typeorm';
import { Field, ID, Int, ObjectType } from 'type-graphql';
import {
  IsNotEmpty,
  IsUUID,
  Max,
  MaxLength,
  Min,
  validateOrReject,
} from 'class-validator';
import Thumbnail from './Thumbnail';

@Entity({ name: 'photos' })
@ObjectType()
export default class Photo {
  @PrimaryColumn('uuid')
  @Field(() => ID)
  @IsUUID('4')
  public readonly id: string;

  @Column({ type: 'bigint' })
  @Field(() => Int)
  @IsNotEmpty()
  @Min(1)
  @Max(10000000000) // 10 gb
  public size: number;

  @Column({ type: 'smallint' })
  @Field(() => Int)
  @IsNotEmpty()
  @Min(1)
  @Max(32767)
  public width: number;

  @Column({ type: 'smallint' })
  @Field(() => Int)
  @IsNotEmpty()
  @Min(1)
  @Max(32767)
  public height: number;

  @Column({ length: 32 })
  @Field()
  @IsNotEmpty()
  @MaxLength(32)
  public mime: string;

  @CreateDateColumn()
  @Field()
  public readonly uploadedTime: Date;

  @OneToMany(() => Thumbnail, thumbnail => thumbnail.photo)
  @Field(() => [Thumbnail])
  public thumbnails: Promise<Thumbnail[]>;

  public constructor(id: string) {
    this.id = id;
  }

  public static getPhotos(): Promise<Photo[]> {
    return getRepository(Photo).find();
  }

  @BeforeInsert()
  @BeforeUpdate()
  public async validate(): Promise<void> {
    await validateOrReject(this);
  }
}
