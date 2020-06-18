import {
  BeforeInsert,
  BeforeUpdate,
  Column,
  CreateDateColumn,
  Entity,
  OneToMany,
  PrimaryColumn,
} from 'typeorm';
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
export default class Photo {
  @PrimaryColumn('uuid')
  @IsUUID('4')
  public readonly id: string;

  @Column({ type: 'bigint' })
  @IsNotEmpty()
  @Min(1)
  @Max(10000000000) // 10 gb
  public size: number;

  @Column({ type: 'smallint' })
  @IsNotEmpty()
  @Min(1)
  @Max(32767)
  public width: number;

  @Column({ type: 'smallint' })
  @IsNotEmpty()
  @Min(1)
  @Max(32767)
  public height: number;

  @Column({ length: 32 })
  @IsNotEmpty()
  @MaxLength(32)
  public mime: string;

  @CreateDateColumn()
  public readonly uploadedTime: Date;

  @OneToMany(() => Thumbnail, thumbnail => thumbnail.photo)
  public thumbnails: Promise<Thumbnail[]>;

  public constructor(id: string) {
    this.id = id;
  }

  @BeforeInsert()
  @BeforeUpdate()
  public async validate(): Promise<void> {
    await validateOrReject(this);
  }
}
